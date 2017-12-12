sudo iptables -t nat -D OUTPUT -p tcp -m tcp -j REDIRECT --dport 443 --to-ports $1 ! --sport $1 ! -s 127.0.0.1
sudo iptables -t nat -D OUTPUT -p tcp -m tcp -j REDIRECT --dport 80 --to-ports $1 ! --sport $1 ! -s 127.0.0.1
sudo iptables -t nat -I OUTPUT -p tcp -m tcp -j REDIRECT --dport 443 --to-ports $1 ! --sport $1 ! -s 127.0.0.1
sudo iptables -t nat -I OUTPUT -p tcp -m tcp -j REDIRECT --dport 80 --to-ports $1 ! --sport $1 ! -s 127.0.0.1

