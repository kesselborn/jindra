TESTS ?= .

.PHONY: testreport.html
testreport.html:
	which go-junit-report
	set -o pipefail; \
	  CONCURRENT=${CONCURRENT} TESTS=${TESTS} make test | tee testreport.log; \
		res=$$?; \
		cat testreport.log | go-junit-report > testreport.xml; \
		./junit2html testreport.xml > testreport.html; \
		test $$res -eq 0

test:
	VERBOSE=${VERBOSE} CONCURRENT=${CONCURRENT} TESTS=${TESTS} ./tests.sh

