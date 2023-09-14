#!/bin/bash

# Determine the user's home directory
USER_HOME=$(eval echo "~$USER")

# Define the name of the virtual environment
VENV_NAME="awst_env"

# Define the path for the virtual environment in the user's home directory
VENV_DIR="$USER_HOME/$VENV_NAME"

echo "SETTING VENV FOR PYTHON ON $VENV_DIR"
# Create the virtual environment if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    python3 -m venv "$VENV_DIR"
fi

# Activate the virtual environment
source "$VENV_DIR/bin/activate"

echo "Installing virtual environment requirements"
# Install project dependencies into the virtual environment
pip install -r requirements.txt


sed -i "" "1s|^|#!$VENV_DIR/bin/python3\n|" awst.py


chmod +x awst.py
cp awst.py /usr/local/bin/awst