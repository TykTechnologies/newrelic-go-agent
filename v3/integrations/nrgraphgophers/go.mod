module github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrgraphgophers

// As of Jan 2020, the graphql-go go.mod file uses 1.13:
// https://github.com/graph-gophers/graphql-go/blob/master/go.mod
go 1.19

require (
	// graphql-go has no tagged releases as of Jan 2020.
	github.com/graph-gophers/graphql-go v1.3.0
	github.com/TykTechnologies/newrelic-go-agent/v3 v3.24.1
)


replace github.com/TykTechnologies/newrelic-go-agent/v3 => ../..
