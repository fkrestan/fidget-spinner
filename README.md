# Fidget Spinner

Fidget Spinner is web service that akin to a real fidget spinners doesn't do any
useful work.

It has a single endpoint `GET /spin` which "spins" CPU for a bit faking useful
work. See it's [OpenAPI doc][oapi].

As surprising as this might be this service has it's uses. It's raison d'être is
testing of both Kubernetes [horizontal pod autoscaling][hpa] and [cluster
capacity autoscaling][ca] and their behavior under various conditions. More
specifically CPU utilization or RPS based autoscaling.


[oapi]: ./api/public.yaml
[hpa]: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/
[ca]: https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler
