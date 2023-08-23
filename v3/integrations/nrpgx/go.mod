module github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrpgx

go 1.19

require (
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.13.0
	github.com/TykTechnologies/newrelic-go-agent/v3 v3.24.1
)


replace github.com/TykTechnologies/newrelic-go-agent/v3 => ../..
