module github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrmssql

go 1.19

require (
	github.com/microsoft/go-mssqldb v0.19.0
	github.com/TykTechnologies/newrelic-go-agent/v3 v3.24.1
)


replace github.com/TykTechnologies/newrelic-go-agent/v3 => ../..
