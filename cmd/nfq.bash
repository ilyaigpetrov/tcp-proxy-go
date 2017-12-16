sudo iptables -D OUTPUT -p tcp -m tcp --dport 80  -j NFQUEUE --queue-num 14 -m owner ! --gid-owner proxyclient
#sudo iptables -D OUTPUT -p tcp -m tcp --dport 443 -j NFQUEUE --queue-num 14 -m owner ! --gid-owner proxyclient
sudo iptables -A OUTPUT -p tcp -m tcp --dport 80  -j NFQUEUE --queue-num 14 -m owner ! --gid-owner proxyclient
#sudo iptables -A OUTPUT -p tcp -m tcp --dport 443 -j NFQUEUE --queue-num 14 -m owner ! --gid-owner proxyclient


