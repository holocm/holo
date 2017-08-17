#!/bin/sh

# try to setup /usr/share/holo/files/03-order/etc/ such that the holoscript is
# found before the plain file (so as to confirm that the repo scanner reorders
ETC=target/usr/share/holo/files/03-order/etc
mv $ETC ${ETC}.old
mkdir $ETC
mv ${ETC}.old/check-ordering.conf.holoscript $ETC
mv ${ETC}.old/check-ordering.conf            $ETC
rmdir ${ETC}.old
