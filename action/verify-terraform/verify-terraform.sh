#!/bin/sh

export GNUPGHOME=./.gnupg

rm -rf ${GNUPGHOME};

gpg --quick-generate-key --batch --passphrase "" github-action@abcxyz.dev

curl --remote-name https://keybase.io/hashicorp/pgp_keys.asc

gpg --import pgp_keys.asc

gpg --sign-key 34365D9472D7468F

gpg --fingerprint --list-signatures "HashiCorp Security"