# Cacheman

![Language](https://img.shields.io/badge/language-Go-blue.svg?style=for-the-badge)[![GitHub license](https://img.shields.io/github/license/athdesk/cacheman?style=for-the-badge)](https://github.com/athdesk/cacheman/blob/master/LICENSE.md)![GitHub repo size](https://img.shields.io/github/repo-size/athdesk/cacheman?style=for-the-badge&color=red)

A tiny pacman centralized caching server, with support for simultaneous downloads!

Installation   
------------
Install from AUR: (not ready yet)
```
$ trizen -S cacheman
$ sudo systemctl enable cacheman --now 
```

Configuration
-----
```
# nano /etc/cacheman.conf 
```

Usage
-----
Set cacheman as the primary pacman mirror on every Arch machine in your network
```
# echo -e "Server = http://CACHEMAN_IP:PORT/$repo/os/$arch\n$(cat /etc/pacman.d/mirrorlist)" > /etc/pacman.d/mirrorlist
```
