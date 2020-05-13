package mdatp

import (
	"fmt"
	"time"
)

var (
	odataDatetimeFormat   = "2006-01-02T15:04:05.99999Z"
	oDataIntervalQueryStr = "%v gt %v and %v le %v"
)

// For now, we do it dirty, but it would not be that hard to have
// a string builder like type with helper methods for creating
// arbitrary OData queries.
func makeIntervalOdataQuery(field string, start, end time.Time) string {
	return fmt.Sprintf(oDataIntervalQueryStr,
		field,
		start.UTC().Format(odataDatetimeFormat),
		field,
		end.UTC().Format(odataDatetimeFormat),
	)
}
