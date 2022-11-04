#!/bin/sh

mkdir -p temp
cd temp || exit

export GNUPGHOME=./.gnupg;

checksum_file=../../terraform-checksums.json

# Generate a temporary key to use for verification
gpg --batch --quick-generate-key --batch --passphrase "" github-action@abcxyz.dev;

# Retrieve the hashicorp key
curl -s --remote-name https://keybase.io/hashicorp/pgp_keys.asc;

# Import the key from hashicorp
gpg --batch --import pgp_keys.asc;

# Sign the hashicorp key with our key
gpg --batch --yes --trust-model always --sign-key 34365D9472D7468F;

release_url=https://releases.hashicorp.com/terraform;

curl -s --remote-name ${release_url}/index.json;

# Exclude all 0.x and pre-release versions
jq -r 'select(.name=="terraform") | .versions[] | select(.version | (contains("-") or startswith("0.")) | not) | .version' < index.json > versions.list;


added_file=added.list;

touch "${added_file}";

while IFS= read -r version; 
do 
    exists=$(jq --arg version "${version}" '.versions[] | select(.version==$version)' < "${checksum_file}");
    if [ "${exists}" = "" ]; 
    then
        version_file=${version}.json;
        jq -r --arg version "${version}" 'select(.name=="terraform") | .versions[] | select(.version==$version)' < index.json > "${version_file}";
        sha_file=$(jq -r '.shasums' < "${version_file}");
        sig_file=$(jq -r '.shasums_signature' < "${version_file}");
        curl -s --remote-name "${release_url}/${version}/${sha_file}";
        curl -s --remote-name "${release_url}/${version}/${sig_file}";

        # Iterate over both supported architectures
        set -- amd64 arm64;
        for arch in "$@";
        do
            bin_url=$(jq -r --arg arch "${arch}" '.builds[] | select(.os=="linux" and .arch==$arch) | .url' < "${version_file}");
            bin_file=$(jq -r --arg arch "${arch}" '.builds[] | select(.os=="linux" and .arch==$arch) | .filename' < "${version_file}");
            # Download the archive, sha file and signature
            curl -s --remote-name "${bin_url}"

            # Verify the signature against the sha file
            gpg --batch --verify "${sig_file}" "${sha_file}";

            # Verify the archive's checksum
            shasum --algorithm 256 --check --ignore-missing "${sha_file}";

            # Extract the binary from the archive
            unzip -qq -o "${bin_file}";

            # Extract only the shasum for the archive we care about
            arch_sum=$(grep "${bin_file}" "${sha_file}" | cut -d' ' -f1);

            # Produce a checksum of the binary
            bin_sum=$(shasum -a 256 terraform | cut -d' ' -f1);

            version_info=$(jq -c -n --arg version "${version}" --arg archive_checksum "${arch_sum}" --arg binary_checksum "${bin_sum}" --arg os "linux" --arg arch "${arch}" '$ARGS.named');
            jq -c --argjson version_info "${version_info}" '.versions += [$version_info]' < "${checksum_file}" > updated.json;
            mv updated.json "${checksum_file}";

        done;

        echo "${version}" >> "${added_file}";
    fi

done < versions.list;

# If there were any changes set some environment variables
if [ -s ${added_file} ]; 
then
    change_count=$(wc -l ${added_file} | tr -s ' ' | cut -d ' ' -f2);
    change_date=$(date +%Y-%m-%d);
    versions=$(cat ${added_file} | tr '\n' ',' | sed 's/,*$//g');

    {
        echo "CHANGES=${change_count}";
        echo "PR_BRANCH=update-checksums-${change_date}";
        echo "UPDATE_DATE=${change_date}";
        echo "VERSIONS=${versions}";
    } >> "${GITHUB_ENV}";
fi;

unset GNUPGHOME;
