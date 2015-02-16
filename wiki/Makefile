all: services.svg services.png

services.svg: services.dot
	dot -Tsvg < services.dot > services.svg

services.png: services.dot
	dot -Tpng < services.dot > services.png
