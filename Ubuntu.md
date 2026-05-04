# Install fpm
gem install fpm

# Build .deb
fpm -s dir \
  -t deb \
  -n topai \
  -v 1.0.0 \
  -p topai_1.0.0_amd64.deb \
  ./topai-linux-x64=/usr/local/bin/topai
