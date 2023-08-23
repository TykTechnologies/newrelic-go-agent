// This sqlx example is a separate module to avoid adding sqlx dependency to the
// nrpgx go.mod file.
module github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrpgx/example/sqlx
go 1.19
require (
	github.com/jmoiron/sqlx v1.2.0
	github.com/TykTechnologies/newrelic-go-agent/v3 v3.24.1
	github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrpgx v0.0.0
)
replace github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrpgx => ../../
replace github.com/TykTechnologies/newrelic-go-agent/v3 => ../../../..
