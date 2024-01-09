resource "helm_release" "signoz" {
  provider          = helm
  name              = "signoz"
  repository        = "https://charts.signoz.io"
  chart             = "signoz"
  namespace         = var.signoz_namespace
  create_namespace  = true
  atomic            = true
  cleanup_on_fail   = true
  dependency_update = true
  set {
    name  = "clickhouse.enabled"
    value = "false"
  }
  set {
    name  = "externalClickhouse.database"
    value = "signoz_main"
  }
  set {
    name  = "externalClickhouse.traceDatabase"
    value = "signoz_traces"
  }
  set {
    name  = "externalClickhouse.secure"
    value = "true"
  }
  set {
    name  = "externalClickhouse.cluster"
    value = "{cluster}"
  }
  set {
    name  = "externalClickhouse.user"
    value = data.doublecloud_clickhouse.backend.connection_info.user
  }
  set {
    name  = "externalClickhouse.user"
    value = data.doublecloud_clickhouse.backend.connection_info.user
  }
  set {
    name  = "externalClickhouse.password"
    value = data.doublecloud_clickhouse.backend.connection_info.password
  }
  set {
    name  = "externalClickhouse.host"
    value = data.doublecloud_clickhouse.backend.connection_info.host
  }
  set {
    name  = "externalClickhouse.httpPort"
    value = data.doublecloud_clickhouse.backend.connection_info.https_port
  }
  set {
    name  = "externalClickhouse.tcpPort"
    value = data.doublecloud_clickhouse.backend.connection_info.tcp_port_secure
  }
}
