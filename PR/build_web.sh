#!/bin/bash

# Build script for Monitor Dashboard Web Interface
set -e

echo "Building Monitor Dashboard Web Interface..."

# Create dist directory if it doesn't exist
mkdir -p /workspace/PR/dist

# Copy web files to dist directory
cp /workspace/PR/web/index.html /workspace/PR/dist/
cp /workspace/PR/web/styles.css /workspace/PR/dist/
cp /workspace/PR/web/app.js /workspace/PR/dist/

echo "Web interface files copied to dist/ directory"

# Create a zip archive of the entire project
cd /workspace/PR
zip -r monitor-dashboard-project.zip . -x "dist/*" "web/*" "*.git/*" "*.gitignore" "build_web.sh" "*.zip"

echo "Project zip archive created: monitor-dashboard-project.zip"

# Also create a zip of just the web interface files
cd web
zip -r ../web-interface.zip .
cd ..

echo "Web interface zip archive created: web-interface.zip"

echo "Build completed successfully!"
echo ""
echo "To use the web interface:"
echo "1. Place the files from the 'dist' directory on your web server or use the existing server's static file serving"
echo "2. Make sure your server config has 'web_dist_dir: \"./dist/\"' pointing to the right location"
echo "3. Access the dashboard at http://your-server:8080"
echo ""
echo "Archives created:"
echo "- monitor-dashboard-project.zip - Complete project"
echo "- web-interface.zip - Just the web interface files"