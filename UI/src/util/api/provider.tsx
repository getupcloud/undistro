import ndjsonStream from 'can-ndjson-stream'

class Provider {
  http: any

  constructor(httpWrapper: any) {
    this.http = httpWrapper
  }

  async list() {
    const url = 'namespaces/undistro-system/clusters/management/proxy/apis/config.undistro.io/v1alpha1/namespaces/undistro-system/providers'
    const res = await this.http.get(url)
    return res.data
  }

  async listMetadata(providerName: string, metadata: string, size: string, page: number, region: string) {
    const url = `/provider/metadata?name=${providerName}&meta=${metadata}&pageSize=${size}&page=${page}&region=${region}`
    const res = await this.http.get(url)
    return res.data
  }

  getEvents() {
    const url = 'http://localhost/uapi/v1/namespaces/undistro-system/clusters/management/proxy/api/v1/namespaces/undistro-system/events?watch=true'
    fetch(url)
      .then((res) => {
        return ndjsonStream(res.body)
      }).then((stream: ReadableStream<Uint8Array> | null) => {
        try {
          let read: any
          stream?.getReader().read().then(read = (result: any) => {
            if (result.done) {
              console.log(result.value)
              return
            }

            console.log(result.value)
            stream.getReader().read().then(read)
          })
        } catch (err) {
          console.log(err)
        }
      })
  }

  async getEvents2() {
    const url = 'http://undistro.local/uapi/v1/namespaces/undistro-system/clusters/management/proxy/api/v1/namespaces/undistro-system/events?watch=true'
    const res = await fetch(url)
    const reader = res.body?.getReader()

    const processText = (stream: ReadableStreamDefaultReadResult<Uint8Array>): Promise<ReadableStreamDefaultReadResult<Uint8Array> | undefined> | undefined => {
      if (stream.done) {
        console.log(stream.value, 'done')
        return
      }

      console.log(stream)
      return reader?.read().then(processText)
    }

    reader?.read().then(processText)
  }

  getEvents3() {
    const url = 'http://undistro.local/uapi/v1/namespaces/undistro-system/clusters/management/proxy/api/v1/namespaces/undistro-system/events?watch=true'
    fetch(url)
      .then((res) => {
        return ndjsonStream(res.body)
      }).then((reader: ReadableStream<Uint8Array> | undefined) => {
        const processText = (stream: ReadableStreamDefaultReadResult<Uint8Array>): Promise<ReadableStreamDefaultReadResult<Uint8Array> | undefined> | undefined => {
          if (stream.done) {
            console.log(stream.value, 'done')
            return
          }

          console.log(reader)
          return reader?.getReader().read().then(processText)
        }

        reader?.getReader().read().then(processText)
      })
  }
}

export default Provider

