# dns

Queries DNS records, currently this resolves directly
to IP addresses rather than CNAMEs etc.

## Supported Methods

* **Get:** A DNS A or AAAA entry to look up
* **List:** **Not supported**
* **Search:** A DNS name (or IP for reverse DNS), this will perform a recursive search and return all results
