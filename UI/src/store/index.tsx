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
  workers: [
    { 
      machineType: string, 
      replicas: number 
    }, 
    { 
      infraNode: boolean,
      machineType: string
    }
  ]
}

const storeWizard = new Store<TypeCreateClusterDefault>({ 
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
  workers: [
    {
      machineType: '',
      replicas: 0
    },
    {
      infraNode: false,
      machineType: ''
    }
  ]
})

export default storeWizard
