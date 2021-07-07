import { Store } from 'pullstate'

type TypeCreateClusterDefault = {
  metadata: {
    name: string,
    namespace: string
  },
  spec: {
    kubernetesVersion: string
    controlPlane: {
      machineType: string,
      replicas: number
    },
    
    infrastructureProvider: {
      flavor: string,
      name: string,
      region: string,
      sshKey: string
    }
  },
  workers: []
}

const storeWizard = new Store<any>({ 
  metadata: {
    name: '',
    namespace: ''
  },
  spec: {
    kubernetesVersion: '',
    controlPlane: {
      machineType: '',
      replicas: 0
    },
    infrastructureProvider: {
      flavor: '',
      name: '',
      region: '',
      sshKey: ''
    }
  },
  workers: []
})

export default storeWizard
