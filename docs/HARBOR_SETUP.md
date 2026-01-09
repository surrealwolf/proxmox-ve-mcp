# Harbor Registry Setup

This project uses Harbor (harbor.dataknife.net) as the container registry for Docker images.

## GitHub Secrets Configuration

To enable automated builds and pushes to Harbor, you need to configure the following GitHub secrets:

### Required Secrets

1. **HARBOR_USER** - Harbor robot account username
   - Value: `robot$library+ci-builder` (replace with your actual robot account username)
   - Location: Repository Settings → Secrets and variables → Actions → New repository secret

2. **HARBOR_PASSWORD** - Harbor robot account password
   - Value: Your Harbor robot account password (set via GitHub CLI or web UI)
   - Location: Repository Settings → Secrets and variables → Actions → New repository secret

### Setting Up GitHub Secrets

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add each secret:
   - Name: `HARBOR_USER`
   - Secret: Your Harbor robot account username (e.g., `robot$library+ci-builder`)
5. Repeat for `HARBOR_PASSWORD` with your Harbor robot account password

**Or use GitHub CLI:**
```bash
gh secret set HARBOR_USER --body "robot\$library+ci-builder"
gh secret set HARBOR_PASSWORD --body "your-password-here"
```

## Local Development

### Manual Login

To manually log in to Harbor:

```bash
# Using make (note: use $$ to escape $ in make)
make docker-login HARBOR_USER='robot$$library+ci-builder' HARBOR_PASSWORD='your-password'
```

Or using docker directly:

```bash
echo 'your-password' | docker login harbor.dataknife.net \
  -u 'robot$library+ci-builder' \
  --password-stdin
```

### Building and Pushing Images

Build the image:

```bash
make docker-build
```

Build and push to Harbor:

```bash
# Note: Use $$ to escape $ character in make
make docker-push HARBOR_USER='robot$$library+ci-builder' HARBOR_PASSWORD='your-password'
```

Pull from Harbor:

```bash
make docker-pull
```

### Using Docker Compose

The `docker-compose.yml` is configured to pull from Harbor by default. To use it:

```bash
# Pull the latest image from Harbor
docker-compose pull

# Run the container
docker-compose up
```

You can override the image location using environment variables:

```bash
HARBOR_REGISTRY=harbor.dataknife.net \
HARBOR_PROJECT=library \
IMAGE_TAG=latest \
docker-compose up
```

## Image Naming Convention

Images are stored in Harbor with the following format:

```
harbor.dataknife.net/library/proxmox-ve-mcp:<tag>
```

Tags are automatically generated based on:
- Branch name (for branch pushes)
- Git SHA (for commits)
- Semantic version (for tags like `v1.0.0`)
- `latest` (for default branch)

## CI/CD Workflow

The GitHub Actions workflow (`.github/workflows/docker-build-push.yml`) automatically:

- Uses **self-hosted runners** for builds
- Builds Docker images on push to main/master branches
- Builds and pushes on version tags (v*)
- Builds (but doesn't push) on pull requests
- Uses Docker layer caching for faster builds
- Tags images with multiple tags for flexibility

**Note:** The workflow is configured to use `runs-on: self-hosted`. Ensure your GitHub repository has self-hosted runners configured and available.

## Harbor Registry Details

- **Registry URL**: `harbor.dataknife.net`
- **Project**: `library`
- **Robot Account**: `robot$library+ci-builder`
- **Image Path**: `harbor.dataknife.net/library/proxmox-ve-mcp`

## Troubleshooting

### Authentication Issues

If you encounter authentication errors:

1. Verify the robot account credentials are correct
2. Check that the robot account has push permissions for the `library` project
3. Ensure GitHub secrets are set correctly (case-sensitive)

### Pull Issues

If pulling images fails:

1. Ensure you're logged in: `make docker-login`
2. Check that the image exists in Harbor
3. Verify the image tag is correct

### Build Cache

The workflow uses Harbor as a build cache. If you need to clear the cache:

1. Delete the `buildcache` tag in Harbor
2. Or wait for it to expire based on Harbor's retention policy
