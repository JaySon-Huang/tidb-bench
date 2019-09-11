host = 127.0.0.1
port = 4000
username = root
scale = 1

tbl: dbgen
	cd dbgen && ./dbgen -v -f -s $(scale)
dbgen:
	cd dbgen; make;
load:
	sh gen_load_script.sh
	mysql -h $(host) -P $(port) -u $(username) -D test < dss.sql
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < CUSTOMER.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < LINEITEM.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < PART.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < NATION.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < ORDERS.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < PARTSUPP.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < REGION.sql&
	nohup mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < SUPPLIER.sql&
watch:
	watch -n 1 "mysql -h $(host) -P $(port) -u $(username) --local-infile=1 -D tpch < alltable.count"

cleansql:
	mv dss.sql dss.sqt
	rm -rf *.sql
	mv dss.sqt dss.sql
cleantbl:
	cd dbgen; make clean; rm -f *.tbl

.PHONY: dbgen load clean tbl
