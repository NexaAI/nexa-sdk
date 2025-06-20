# Follow https://royzsec.medium.com/install-go-1-21-0-in-ubuntu-22-04-2-in-5-minutes-468a5330c64e

# 1. Update package lists
sudo apt-get update

# 2. Check if Go tarball exists, if not download it
if [ ! -f "go1.21.0.linux-amd64.tar.gz" ]; then
    echo "Downloading Go 1.21.0 tarball..."
    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
else
    echo "Go tarball already exists, using local copy..."
fi

# 3. Remove any previous Go installation (optional but recommended)
sudo rm -rf /usr/local/go

# 4. Extract the tarball to /usr/local
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 5. Set proper permissions for /usr/local/go directory
sudo chown -R root:root /usr/local/go
sudo chmod -R 755 /usr/local/go

# 6. Set Go environment variables (add these to ~/.profile or ~/.bashrc for persistence)
echo 'export GOROOT=/usr/local/go' >> ~/.profile
echo 'export GOPATH=$HOME/go' >> ~/.profile
echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.profile

# 7. Also add to current session
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# 8. Create GOPATH directory if it doesn't exist
mkdir -p $HOME/go/{bin,src,pkg}

# 9. Reload profile to apply changes
source ~/.profile

# 10. Verify Go installation and access
go version
ls -la /usr/local/go/bin/go