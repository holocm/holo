This test checks provisioning of SSH keys in various scenarios:

* `user1` does not have any authorized keys, and one key is provisioned.
* `user2` does not have any authorized keys, and a set of two keys is provisioned.
* `user3` does not have any authorized keys, and two sets of keys are provisioned.
* `user4` has some authorized keys, and one key is provisioned.
* `user5` has some authorized keys, and a set of two keys is provisioned.
* `user6` has some authorized keys, and two sets of keys are provisioned.

Also, as a special case, `user7` has some authorized keys, and a set of keys is
provisioned that contains keys that are also contained in the key set that is
provisioned. The expected behavior is that Holo does not touch the other keys at all.

Another special case, `user8` has two overlapping sets of keys. The expected
behavior is that Holo provisions both, since keysets should not interfere with
each other.
