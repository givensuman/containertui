#!/bin/bash
set -e

# Create symbolic links for easier access
if [ ! -L /usr/bin/containertui ]; then
	ln -s /usr/local/bin/containertui /usr/bin/containertui || true
fi

echo "containertui installed successfully!"
