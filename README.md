# Pon Stick Exporter

用于监控 ODI 猫棒状态

![Grafana监控](./dist/grafana.png)

## 使用

1. 启动服务

```shell
docker compose up -d
```

2. 添加 Prometheus 任务

```
- job_name: pon-stick-exporter
  honor_timestamps: true
  scrape_interval: 15s
  scrape_timeout: 10s
  metrics_path: /metrics
  scheme: http
  static_configs:
  - targets:
    - pon-stick-exporter:9001
```

3. 导入 Grafana 面板

导入 [./dist/panel.json](./dist/panel.json) 文件