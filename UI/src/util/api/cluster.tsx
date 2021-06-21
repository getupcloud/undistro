class Cluster {
  http: any

  constructor (httpWrapper: any) {
    this.http = httpWrapper
  }

  async list (namespace: string) {
    const url = `namespaces/undistro-system/clusters/management/proxy/apis/app.undistro.io/v1alpha1/namespaces/${namespace}/clusters`
    const res = await this.http.get(url)
    return res.data
  }

  async post (data: {}, namespace: string) {
    const url = `namespaces/undistro-system/clusters/management/proxy/apis/app.undistro.io/v1alpha1/namespaces/${namespace}/clusters`
    const res = await this.http.post(url, data)
    return res.data
  }

  async postPolicies (data: {}, namespace: string) {
    const url = `namespaces/undistro-system/clusters/management/proxy/apis/app.undistro.io/v1alpha1/namespaces/${namespace}/defaultpolicies`
    const res = await this.http.post(url, data)
    return res.data
  }
}

export default Cluster