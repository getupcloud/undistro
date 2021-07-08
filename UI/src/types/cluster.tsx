export type TypeWorker = {
  id: string,
  machineType: TypeOption | null,
  replicas: number,
  infraNode: boolean
}

export type TypeOption = {
  value: string,
  label: string,
}

export type TypeSelectOptions = {
  [instanceType: string]: {
    selectOptions: TypeOption[];
  };
};

export type TypeAsyncSelect = {
  label?: string,
  onChange: (option: TypeOption | null) => void,
  loadOptions: any,
  value: TypeOption | null
}

export type TypeInfra = {
  provider: string
  setProvider: Function
  providerOptions: any
  regionOptions: TypeOption[]
  region: string
  setRegion: Function
  flavor: string
  setFlavor: Function
  flavorOptions: TypeOption[]
  k8sVersion: string
  setK8sVersion: Function
  sshKey: string
  setSshKey: Function,
  k8sOptions: any,
  sshKeyOptions: string[]
}

export type TypeControlPlane = {
  replicas: number
  setReplicas: Function
  cpu: TypeOption | null
  setCpu: Function
  getCpu: Function
  getMem: Function
  memory: TypeOption | null
  setMemory: Function
  machineTypes: TypeOption | null
  setMachineTypes: Function
  getMachineTypes: Function
  infraNode: boolean
  setInfraNode: Function
  replicasWorkers: number
  setReplicasWorkers: Function
  cpuWorkers: TypeOption | null
  setCpuWorkers: Function
  memoryWorkers: TypeOption | null
  setMemoryWorkers: Function
  machineTypesWorkers: TypeOption | null
  setMachineTypesWorkers: Function
  createWorkers: () => void
  workers: TypeWorker[]
  deleteWorkers: Function
  clusterName: string
}