class Node {
  http: any

  constructor (httpWrapper: any) {
    this.http = httpWrapper
  }

  async list () {
    const url = '/api/v1/nodes'
    const res = await this.http.get(url)
    return res.data
  }
}

export default Node