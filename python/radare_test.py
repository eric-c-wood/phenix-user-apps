#!/usr/bin/python3.10

import r2pipe
import sys
import hashlib

def main():
	r2 = r2pipe.open(sys.argv[1])
	get_imported_called_functions(r2)
	list_dependencies(r2)
	#get_binary_info(r2)
	#get_file_hashes(r2)
	#get_exports(r2)
	get_file_hash(sys.argv[1],"md5")
	get_file_hash(sys.argv[1],"sha256")
	
def get_exports(r2):
	json_import = r2.cmd('iEj')
	print(json_import)
	
def get_md5sum(filepath):
	with open(filepath,'rb') as input:
		md5_hash = hashlib.md5()
		chunk = input.read(8192)
		while chunk:
			md5_hash.update(chunk)
			chunk = input.read(8192)
		
		print("MD5:{}".format(md5_hash.hexdigest()))
		
def get_file_hash(filepath,algorithm="sha256"):
	with open(filepath,'rb') as input:
		if "sha256" in algorithm:
			file_hash = hashlib.sha256()
		else:
			file_hash = hashlib.md5()
			
		chunk = input.read(8192)
		while chunk:
			file_hash.update(chunk)
			chunk = input.read(8192)
		
		print("{}sum:{}".format(algorithm,file_hash.hexdigest()))
		
	
def list_dependencies(r2):
	json_import = r2.cmdj('ilj')
	for item in json_import:
		print(item)
		
def get_binary_info(r2):
	json_import = r2.cmdj('iIj')
	print(json_import)
		
def get_imported_called_functions(r2):
	#r2.cmd('aaa')
	json_import = r2.cmdj('iij')
	for item in json_import:
		if len(item.get("name","")) == 0:
			continue
		#xref = r2.cmd('aaf;axt @@ *{}*'.format(item.get("name","")))
		
		print("Library:{} Name:{} Called:{}".format(item.get("libname",""),item.get("name",""),True))
					

if __name__ == '__main__':
	print("Paths:{}".format(sys.path))
	print("Arguments:{}".format(sys.argv))
	main()
