## X.X.X / YYYY-MM-DD

* [CHANGE] Use libc for GH binary releases and muslc for docker image
* [CHANGE] Upgrade all vendored libraries
* [CHANGE] Now the build of binaries (docker & standalone) is dynamically linked and not
    static, this is required by the dynamic plugin load feature
* [FEATURE] Add plugins system to extend Ladder blocks
* [CHANGE]  Upgrade to go 1.8

## v0.1.0 / 2017-05-05

* [FEATURE] Autoscalers logic
* [FEATURE] Gatherers: cloudwatch, sqs, random, prometheus
* [FEATURE] Arrangers: threshold, constantFactor
* [FEATURE] Solvers: bound
* [FEATURE] Filters: ecsRunningTasks, limit, scalingKindInterval
* [FEATURE] Scalers: AWS ASG, AWS ECS
* [FEATURE] API endpoints: autoscalerList, stopAutoscaler, cancelStopAutoscaler
* [FEATURE] Documentation
* [FEATURE] Prometheus metrics
* [FEATURE] Health check
