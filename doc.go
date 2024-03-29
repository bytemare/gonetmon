/*
Package gonetmon is a HTTP traffic monitor, capturing packets on all open interfaces and presenting results to the console.

It uses gopacket to sniff traffic packets and, based on filters, allows to select allowed packets and analyse them further.
gonetmon's main features are :

	* a display giving the operator real-time insight about the traffic
	* the number of total http packets received over a specified time frame
	* current traffic speed
	* network interfaces used by the traffic
	* the most visited website over a specified time frame, sections visited, request methods and response codes
	* alerting whenever the traffic hits a defined threshold, and when it recovered.

The project contains a ready-to-use monitor to start checking out traffic.

*/
package gonetmon
