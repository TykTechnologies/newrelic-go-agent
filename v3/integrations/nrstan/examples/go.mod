module github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrstan/examples
// This module exists to avoid a dependency on nrnrats.
go 1.19
require (
	github.com/nats-io/stan.go v0.5.0
	github.com/TykTechnologies/newrelic-go-agent/v3 v3.24.1
	github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrnats v0.0.0
	github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrstan v0.0.0
)
replace github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrstan => ../
replace github.com/TykTechnologies/newrelic-go-agent/v3/integrations/nrnats => ../../nrnats/
replace github.com/TykTechnologies/newrelic-go-agent/v3 => ../../..
