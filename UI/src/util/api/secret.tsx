class Secret {
  http: any

  constructor (httpWrapper: any) {
    this.http = httpWrapper
  }

  async list () {
    const url = '/api/v1/namespaces/undistro-system/secrets/undistro-aws-config'
    const res = await this.http.get(url)
    return res.data
  }
}

export default Secret