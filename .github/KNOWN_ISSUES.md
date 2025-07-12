# Known Issues

## GitHub Actions

### Cache Warnings
- **Issue**: "Cannot open: File exists" errors during cache restoration
- **Impact**: None - tests run successfully
- **Status**: GitHub Actions known issue with concurrent cache access
- **Solution**: Added `continue-on-error: true` to cache steps

### Windows Path Validation
- **Issue**: "Path Validation Error" on Windows cache steps
- **Impact**: None - cache is optional
- **Status**: Fixed by separating Unix/Windows cache logic

### macOS Runner Migration Notice
- **Issue**: "The macos-latest label will migrate to macOS 15"
- **Impact**: None - informational notice
- **Date**: August 4, 2025
- **Action Required**: None - code is compatible with macOS 15