On running the command. All errors MUST be resolved otherwise import files may not do whats expected.

Main Command example:

```bash
  ./apoxfrm -config-file root.yaml -extnet-prefix customer:ext:net
```

Other Examples:

```bash
  ./apoxfrm -config-file zone.yaml -extnet-prefix customer:ext:net -extra-files root.yaml
  ./apoxfrm -config-file tenant-a.yaml -extnet-prefix customer:ext:net -extra-files root.yaml zone.yaml
  ./apoxfrm -config-file tenant-b.yaml -extnet-prefix customer:ext:net -extra-files root.yaml zone.yaml
  ./apoxfrm -config-file tenant-c.yaml -extnet-prefix customer:ext:net -extra-files root.yaml zone.yaml
```
