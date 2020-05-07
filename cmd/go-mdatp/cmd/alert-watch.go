package cmd

import (
	"context"
	"errors"
	"fmt"
	"go-mdatp/pkg/mdatp"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	errInvalidStatefile = errors.New("statefile content empty or invalid, starting fresh")
)

type configAlertWatch struct {
	ConfigFile string

	LogFile   string
	StateFile string

	Output string
	Indent bool

	Debug       bool
	JSONLogging bool
}

func setupCmdAlertWatch(cmd *cobra.Command, c *configAlertWatch) *cobra.Command {
	envconfig.Process("", c)
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&c.ConfigFile, "config", "c", c.ConfigFile, "Set configfile alternate location. Default is $CWD/.go-mdatp.yaml.")

	cmd.Flags().StringVarP(&c.LogFile, "log", "l", c.LogFile, "Set logging output to provided file. Default is stderr.")
	cmd.Flags().StringVarP(&c.StateFile, "state", "s", c.StateFile, "Set state output to provided file. Default is to not persist state.")

	cmd.Flags().StringVarP(&c.Output, "output", "o", c.Output, "Set records output. Available schemes: file://path/to/file, udp://1.2.3.4:1234, tcp://1.2.3.4:1234")
	cmd.Flags().BoolVarP(&c.Indent, "indent", "i", c.Indent, "Set records output to be indented.")

	cmd.Flags().BoolVarP(&c.Debug, "debug", "d", c.Debug, "Set log level to DEBUG.")
	cmd.Flags().BoolVar(&c.JSONLogging, "json", c.JSONLogging, "Set log formatter to JSON.")
	return cmd
}

func newCommandWatch() *cobra.Command {
	var cmdCfg configAlertWatch
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Query audit records at regular intervals.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := initConfig(cmdCfg.ConfigFile)
			if err != nil {
				return err
			}
			logger, err := initLogger(cmd, cmdCfg.LogFile, cmdCfg.Debug, cmdCfg.JSONLogging)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				sigChan := getSigChan()
				for {
					select {
					case <-sigChan:
						cancel()
						return
					}
				}
			}()

			rwc, err := setupOutput(ctx, cmdCfg.Output)
			if err != nil {
				return err
			}
			if cmdCfg.Output != "" {
				logger.Infof("using output: %s", cmdCfg.Output)
			}

			var hasStateSource bool
			var stateRwc io.ReadWriteCloser
			if cmdCfg.StateFile != "" {
				if stateRwc, err = os.OpenFile(cmdCfg.StateFile, os.O_RDWR|os.O_CREATE, 0640); err != nil {
					return err
				}
				hasStateSource = true
			}

			client, err := mdatp.NewClient(
				mdatp.WithOAuthClient(
					config.Credentials.ClientID,
					config.Credentials.ClientSecret,
					config.Credentials.TenantID,
				),
			)
			if err != nil {
				return err
			}

			req := &mdatp.AlertWatchRequest{
				OutputSource:   rwc,
				IsOutputIndent: cmdCfg.Indent,
				HasStateSource: hasStateSource,
				StateSource:    stateRwc,
				Logger:         logger,
				State:          mdatp.NewWatchStateJSON(),
			}
			if err := client.Alert.Watch(ctx, req); err != nil {
				return err
			}
			return nil
		},
	}
	return setupCmdAlertWatch(cmd, &cmdCfg)
}

func getSigChan() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	return sigChan
}

func initLogger(cmd *cobra.Command, logFile string, setDebug, setJSON bool) (*logrus.Logger, error) {
	logger := logrus.New()

	logger.SetLevel(logrus.InfoLevel)
	if setDebug {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableSorting:         true,
	})
	if setJSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger.SetOutput(loggerOutput)
	if logFile != "" {
		logFile, err := filepath.Abs(logFile)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute filepath for provided logfile: %s", err)
		}
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			return nil, fmt.Errorf("could not use provided logfile: %s", err)
		}
		logger.SetOutput(f)
		cmd.PostRunE = func(cmd *cobra.Command, args []string) error {
			return f.Close()
		}
	}
	return logger, nil
}

func setupOutput(ctx context.Context, selection string) (io.ReadWriteCloser, error) {
	var err error
	var rwc io.ReadWriteCloser

	filePrefix := "file://"
	udpPrefix := "udp://"
	tcpPrefix := "tcp://"

	netDial := func(scheme, addr string) (io.ReadWriteCloser, error) {
		var d net.Dialer
		return d.DialContext(ctx, scheme, addr)
	}

	switch {
	default:
		return nil, fmt.Errorf("output invalid")
	case selection == "":
		rwc = defaultOutput
	case strings.HasPrefix(selection, filePrefix):
		path := strings.TrimPrefix(selection, filePrefix)
		path, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute filepath for provided statefile: %s", err)
		}
		rwc, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	case strings.HasPrefix(selection, udpPrefix):
		path := strings.TrimPrefix(selection, filePrefix)
		rwc, err = netDial("udp", path)
	case strings.HasPrefix(selection, tcpPrefix):
		path := strings.TrimPrefix(selection, filePrefix)
		rwc, err = netDial("tcp", path)
	}
	return rwc, err
}
