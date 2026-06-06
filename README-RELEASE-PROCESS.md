# Release Process (Manual)

We currently do releases manually. Automation will be added with GitHub Actions.

Tracking issue:
<https://github.com/iam404/endpoint-monitoring-operator/issues/14>

## Checklist

Use this checklist exactly for each release.

### 1. Prepare Local Repo

```bash
git checkout main
git pull origin main
git status
```

Make sure the working tree is clean before releasing.

### 2. Set Release Variables

```bash
export VERSION=v1.0.2
export IMG=docker.io/tarunteja417/endpoint-monitoring-operator:$VERSION
```

### 3. Login to Docker Hub

```bash
docker login
```

### 4. Build Image From Latest Code

```bash
make docker-build IMG=$IMG
```

### 5. Security Gate (Secrets Only)

Run secret scans before push.

```bash
trivy fs --scanners secret --skip-dirs .git .
trivy image --scanners secret $IMG
```

If either scan reports sensitive content, stop and fix before continuing.

### 6. Push Release Image

```bash
make docker-push IMG=$IMG
```

### 7. Regenerate Installer With New Image

```bash
make build-installer IMG=$IMG
```

This updates:

- `dist/install.yaml`

### 8. Verify Installer Points to Release Image

```bash
rg -n "image:" dist/install.yaml | head -n 5
```

Expected image value:

`docker.io/tarunteja417/endpoint-monitoring-operator:$VERSION`

### 9. Commit, Tag, and Push

```bash
git add dist/install.yaml
git commit -m "release: $VERSION"
git tag -a $VERSION -m "Release $VERSION"
git push origin main
git push origin $VERSION
```

### 10. Publish and Share Install Command

Users install that exact release with:

```bash
kubectl apply -f https://raw.githubusercontent.com/iam404/endpoint-monitoring-operator/$VERSION/dist/install.yaml
```

### Rollback

If a release is bad, publish a new patch version (for example, `v1.0.3`) from a fixed commit.
Do not reuse or overwrite an existing release tag.
