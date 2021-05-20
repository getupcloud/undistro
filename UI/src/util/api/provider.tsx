class Provider {
  http: any

  constructor (httpWrapper: any) {
    this.http = httpWrapper
  }

  async list () {
    const url = '/apis/config.undistro.io/v1alpha1/namespaces/undistro-system/providers'
    const res = await this.http.get(url)
    return res.data
  }
}

export default Provider