# pdf-report-demo
Service for generating reports in [PDF](http://pdf.io) format from [JSON](http://json.org) object stored in [Redis](http://redis.io/)

# Build & Run
Download the sources from the repository and run the service in the docker container.

*Before starting, make sure you have free ports 8080 and 6379*

```
git clone https://github.com/pshvedko/pdf-report.git
cd pdf-report
docker-compose up
```

# Usage
Open http://localhost:8080 in browser
