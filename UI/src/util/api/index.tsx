import axios from 'axios'
import Node from './node'
import Cluster from './cluster'
import Provider from './provider'
import Secret from './secret'

const HOST = 'localhost'

const BASE_URL = `http://${HOST}/uapi/v1`

const httpWrapper = axios.create({
  baseURL: BASE_URL + '/',
  timeout: 600000
})

const Api = {
  Node: new Node(httpWrapper),
  Cluster: new Cluster(httpWrapper),
  Provider: new Provider(httpWrapper),
  Secret: new Secret(httpWrapper)
}

export default Api