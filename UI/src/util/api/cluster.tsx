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

  async post (data: {}) {
    const url = '/apis/app.undistro.io/v1alpha1/namespaces/default/clusters'
    const res = await this.http.post(url, data)
    return res.data
  }

  async postPolicies (data: {}) {
    const url = '/apis/app.undistro.io/v1alpha1/namespaces/default/defaultPolicies'
    const res = await this.http.post(url, data)
    return res.data
  }
}

export default Cluster