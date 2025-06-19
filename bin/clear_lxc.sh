for a in `find |grep loxcache|grep lxc$`
do
	echo $a
	rm $a
done
