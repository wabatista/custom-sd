# custom-sd
Custom query based on prometheus promql, to make a instant query on prometheus api and return de match instances as a file_sd.

### Query:
```bash
role{role=~'%s', exporter_port=~'.+', metrics_path=~'.+', app=~'.+'}
```
#### With that, we can generate customized files_sd for other exporters using existing metrics as a custom.prom in node_exporter
```bash 
role.prom on node_exporter's text.dir :
role{role=~'jmx_exporter', exporter_port=~'9999', metrics_path=~'/prometheus/metrics', app=~'application_a'}
```

### Result:
```bash 
jmx_exporter.metrics.json
[
        "targets": [
            "instance_address:9999"
        ],
        "labels": {
            "__meta_app": "application_a",
            "__meta_exporter_port": "9999",
            "__meta_instance": "instance_address:9100", 
            "__meta_ip": "__address__",
            "__meta_job": "__job__",
            "__meta_metrics_path": "/metrics",
            "__meta_role": "jmx_exporter",
            "__name__": "role"
        }
    }
]
```
#### Node: The instance_address will be generate by labels existents on node_exporter's metrics
