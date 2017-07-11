#!/usr/bin/python

import argparse
import re
import sys
import datetime
import time
import os
import sys

reconPath = "/usr/share/recon-ng/"

sys.path.insert(0,reconPath)
from recon.core import base
from recon.core.framework import Colors

def run_module(reconBase, module, domain):
    x = reconBase.do_load(module)
    x.do_set("SOURCE " + domain)
    x.do_run(None)

def run_recon(domains, output, shodankey):
	stamp = datetime.datetime.now().strftime('%M:%H-%m_%d_%Y')
	wspace = domains[0]+stamp

	reconb = base.Recon(base.Mode.CLI)
	reconb.init_workspace(wspace)
	reconb.onecmd("TIMEOUT=100")
	reconb.onecmd("keys add shodan_api " + shodankey)

	module_list = [
		"recon/domains-hosts/threatcrowd",
		"recon/domains-hosts/hackertarget",
		"recon/domains-hosts/bing_domain_web",
		"recon/domains-hosts/shodan_hostname",
	]

	for domain in domains:
		for module in module_list:
	    		run_module(reconb, module, domain)

	x = reconb.do_load("reporting/list")
	x.do_set("FILENAME " + output)
	x.do_set("COLUMN host")
	x.do_run(None)

parser = argparse.ArgumentParser()
parser.add_argument("-o", dest="output", help="output file", default=None)
parser.add_argument("-s", dest="shodankey", help="shodan key", default=None)
parser.add_argument("domains", help="one or more domains", nargs="*", default=None)
args = parser.parse_args()

domainList = []

if args.domains:
 	domainList+=args.domains

run_recon(domainList, args.output, args.shodankey)
