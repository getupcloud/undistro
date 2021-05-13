class Cluster {
  http: any

  constructor (httpWrapper: any) {
    this.http = httpWrapper
  }

  async list () {
    const url = '/apis/app.undistro.io/v1alpha1/namespaces/default/clusters'
    const res = await this.http.get(url)
    return res.data
  }
}

export default Cluster