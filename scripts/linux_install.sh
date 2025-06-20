# Follow https://royzsec.medium.com/install-go-1-21-0-in-ubuntu-22-04-2-in-5-minutes-468a5330c64e

# 1. Update package lists
sudo apt-get update

# 2. Download Go 1.21 tarball
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz

# 3. Remove any previous Go installation (optional but recommended)
sudo rm -rf /usr/local/go

# 4. Extract the tarball to /usr/local
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 5. Set Go environment variables (add these to ~/.profile or ~/.bashrc for persistence)
echo 'export GOROOT=/usr/local/go' >> ~/.profile
echo 'export GOPATH=$HOME/go' >> ~/.profile
echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.profile

# 6. Reload profile to apply changes
source ~/.profile

# 7. Verify Go installation
go version