val majorMinor = "1.0"
val buildNumber = {{ .BuildNumber }}

version := {
  if ({{ .Snapshot }})
    s"$majorMinor.$buildNumber-SNAPSHOT"
  else
    s"$majorMinor.$buildNumber"
}
