import axios from 'axios'
import Node from './node'
import Cluster from './cluster'
import Provider from './provider'
import Secret from './secret'

const BASE_URL = 'http://localhost:2020/uapi/v1/namespaces/undistro-system/clusters/management/proxy'

const httpWrapper = axios.create({
  baseURL: BASE_URL + '/',
  timeout: 10000
})

const Api = {
  Node: new Node(httpWrapper),
  Cluster: new Cluster(httpWrapper),
  Provider: new Provider(httpWrapper),
  Secret: new Secret(httpWrapper)
}

export default Api