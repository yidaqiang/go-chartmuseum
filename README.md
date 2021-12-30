# go-chartmuseum
go library for chartmuseum


## TODO

### Chartmuseum API 

------

Helm Chart Repository

- [ ] `GET /index.yaml`  - retrieved when you run `helm repo add chartmuseum http://localhost:8080/`
- [ ] `GET /charts/mychart-0.1.0.tgz`  retrieved when you run `helm install chartmuseum/mychart`
- [ ] `GET /charts/mychart-0.1.0.tgz.prov`  - retrieved when you run `helm install` with the `--verify flag`

Chart Manipulation

- [x] `POST /api/charts` - upload a new chart version
- [ ] `POST /api/prov` - upload a new provenance file
- [x] `DELETE /api/charts/<name>/<version>` - delete a chart version (and corresponding provenance file)
- [x] `GET /api/charts` - list all charts
- [x] `GET /api/charts/<name>` - list all versions of a chart
- [x] `GET /api/charts/<name>/<version>` - describe a chart version
- [x] `HEAD /api/charts/<name>` - check if chart exists (any versions)
- [x] `HEAD /api/charts/<name>/<version>` - check if chart version exists

Server Info

- [ ] `GET /` - HTML welcome page
- [ ] `GET /info` - returns current ChartMuseum version
- [ ] `GET /health` - returns 200 OK
