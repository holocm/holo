This testcase validates the automatic migration from
`/var/lib/holo/users-groups/state.toml` (used by versions 1.2 and 1.3) to the
new base images and provisioned images. The following situation is simulated:

* `group:created` was created by Holo, so the migration will write an empty
  base image and create the appropriate provisioned image.

* `user:modified` was modified by Holo, and the system administrator already
  provided the appropriate base image before upgrading holo-users-groups. This
  image will be used and only the provisioned image will be created.

Both entities were provisioned before the upgrade, so they are both noted in
the `state.toml`.
