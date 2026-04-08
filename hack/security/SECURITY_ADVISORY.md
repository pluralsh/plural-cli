# Security Advisory: OpenSSL Vulnerability Fix

## Issue Summary

**Vulnerability**: OpenSSL Denial of Service via malformed PKCS#12 file processing  
**Severity**: Low  
**Affected Package**: libssl3  
**Vulnerable Version**: 3.5.4-r0  
**Fixed Version**: 3.5.5-r0

## Description

Processing a malformed PKCS#12 file can trigger a NULL pointer dereference in the `PKCS12_item_decrypt_d2i_ex()` function, leading to a crash and Denial of Service for applications processing PKCS#12 files.

The `PKCS12_item_decrypt_d2i_ex()` function does not check whether the `oct` parameter is NULL before dereferencing it. When called from `PKCS12_unpack_p7encdata()` with a malformed PKCS#12 file, this parameter can be NULL, causing a crash.

**Impact**: The vulnerability is limited to Denial of Service and cannot be escalated to achieve code execution or memory disclosure.

## Resolution

### Docker Images

For any deployments using `curlimages/curl:latest`, the image has been pinned to a secure digest:

```
curlimages/curl@sha256:b066cbf876d50a5d024927878a586c4a39c985325ded195e2e231f4abdddf3c8
```

This digest corresponds to an image with:
- curl 8.19.0
- OpenSSL 3.5.5 (fixed version)
- libssl3 3.5.5-r0

### Alpine-based Images

All Alpine-based images in this repository (golang:1.26.1-alpine3.22) already use libssl3 3.5.5-r0 and are not affected by this vulnerability.

## Verification

To verify the OpenSSL version in a running container:

```bash
# Check libssl3 package version
apk info libssl3

# Check OpenSSL version
openssl version
```

## References

- OpenSSL Security Advisory
- Alpine Linux Package: libssl3
- Fix Version: 3.5.5-r0

## Affected OpenSSL Versions

- OpenSSL 3.6
- OpenSSL 3.5
- OpenSSL 3.4
- OpenSSL 3.3
- OpenSSL 3.0
- OpenSSL 1.1.1
- OpenSSL 1.0.2

**Note**: The FIPS modules in 3.6, 3.5, 3.4, 3.3 and 3.0 are not affected by this issue, as the PKCS#12 implementation is outside the OpenSSL FIPS module boundary.

## Timeline

- **2026-04-08**: Vulnerability identified and fixed version implemented
- Image digest pinned to secure version with libssl3 3.5.5-r0
