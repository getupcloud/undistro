class Provider {
  http: any

  constructor (httpWrapper: any) {
    this.http = httpWrapper
  }

  async list () {
    const url = 'namespaces/undistro-system/clusters/management/proxy/apis/config.undistro.io/v1alpha1/namespaces/undistro-system/providers'
    const res = await this.http.get(url)
    return res.data
  }

  async listMetadata (providerName: string, metadata: string, size: string, page: number) {
    const url = `/provider/metadata?name=${providerName}&meta=${metadata}&page_size=${size}&page=${page}`
    const res = await this.http.get(url)
    return res.data
  }
 
  async getEvents () {
    const url = 'namespaces/undistro-system/clusters/management/proxy/api/v1/namespaces/undistro-system/events?watch=true'
    const res = await this.http.get(url)
    return res.data 
  }
}

export default Provider