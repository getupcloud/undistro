/* eslint-disable react-hooks/exhaustive-deps */
import React, { FC, useEffect, useState } from 'react'
import store from '../store'
import Input from '@components/input'
import Select from '@components/select'
import AsyncSelect, { OptionType } from '@components/asyncSelect'
import { LoadOptions } from "react-select-async-paginate";
// import Modals from 'util/modals'
import { generateId } from 'util/helpers'
import Steps from './steps'
import Button from '@components/button'
import Api from 'util/api'
import Toggle from '@components/toggle'

type Props = {
  handleClose: () => void
}

type TypeWorker = {
  id: string,
  machineType: OptionType | null, 
  replicas: number,
  infraNode: boolean
}

type Option = {
  value: string,
  label: string,
}

type newStruct = {
  [instanceType: string]: {
    selectOptions: Option[];
  };
};


const ClusterWizard: FC<Props> = ({ handleClose }) => {
  const body = store.useState((s: any) => s.body)
  const [accessKey, setAccesskey] = useState<string>('')
  const [secret, setSecret] = useState<string>('')
  const [region, setRegion] = useState<string>('')
  const [clusterName, setClusterName] = useState<string>('')
  const [namespace, setNamespace] = useState<string>('')
  const [provider, setProvider] = useState<string>('')
  const [flavor, setFlavor] = useState<string>('')
  const [k8sVersion, setK8sVersion] = useState<string>('')
  const [replicas, setReplicas] = useState<number>(0)
  const [infraNode, setInfraNode] = useState<boolean>(false)
  const [workers, setWorkers] = useState<TypeWorker[]>([])
  const [machineTypes, setMachineTypes] = useState<OptionType | null>(null)
  const [memory, setMemory] = useState<OptionType | null>(null)
  const [cpu, setCpu] = useState<OptionType | null>(null)
  const [replicasWorkers, setReplicasWorkers] = useState<number>(0)
  const [memoryWorkers, setMemoryWorkers] = useState<OptionType | null>(null)
  const [cpuWorkers, setCpuWorkers] = useState<OptionType | null>(null)
  const [machineTypesWorkers, setMachineTypesWorkers] = useState<OptionType | null>(null)
  const [regionOptions, setRegionOptions] = useState<[]>([])
  const [flavorOptions, setFlavorOptions] = useState<Option[]>([])
  const [k8sOptions, setK8sOptions] = useState<newStruct>()
  const [sshKey, setSshKey] = useState<string>('')
  const [sshKeyOptions, setSshKeyOptions] = useState<string[]>([])
  const providerOptions = [{ value: provider, label: 'aws' }]

  const handleAction = () => {
    const getWorkers = workers.map(elm => ({ machineType: elm.machineType, replicas: elm.replicas, infraNode: elm.infraNode  }))

    const cluster = {
      "name": clusterName,
      "namespace": namespace
    }

    const spec = {
      "kubernetesVersion": k8sVersion,
      "controlPlane": {
        "machineType": machineTypes,
        "replicas": replicas
      },

      "infrastructureProvider": {
        "flavor": flavor,
        "name": provider,
        "region": region,
        "sshKey": sshKey
      },

      "workers": getWorkers
    }

    const data = {
      "apiVersion": "app.undistro.io/v1alpha1",
      "kind": "Cluster",
      "metadata": cluster,
      "spec": spec
    }

    const dataPolicies = {
      "apiVersion": "app.undistro.io/v1alpha1",
      "kind": "DefaultPolicies",
      "metadata": {
        "name": "defaultpolicies-undistro",
        "namespace": namespace
      },
      "spec": {
        "clusterName": clusterName
      }
    }

    Api.Cluster.post(data, namespace)
    
    setTimeout(() => {
      Api.Cluster.postPolicies(dataPolicies, namespace)
    }, 600)
  }

  const getSecrets = () => {
    Api.Secret.list()
      .then(res => {
        setAccesskey(atob(res.data.accessKeyID))
        setSecret(atob(res.data.secretAccessKey))
        setRegion(atob(res.data.region))
      })
  }

  const createWorkers = () => {
    setWorkers([...workers, 
      {
        id: generateId(),
        machineType: machineTypesWorkers, 
        replicas: replicasWorkers,
        infraNode: infraNode
      }
    ])
  }

  const deleteWorkers = (id: any) => {
    const newWorkers = workers.filter(item => item.id !== id)
    setWorkers(newWorkers)
  }

  const getProviders = () => {
    Api.Provider.list()
      .then(res => {
        setProvider(res.items[0].metadata.name)
      })
  }

  const getMachineTypes: LoadOptions<OptionType, { page: number }> = async (value, loadedOptions, additional: any) => {
    const res = await Api.Provider.listMetadata('aws', 'machine_types', '15', additional ? additional.page : 1, region)
    const totalPages = res.TotalPages
    const filteredMachineTypes = res.MachineTypes.filter((elm: any) => elm.availability_zones === region)
    const machineTypes = res.MachineTypes.map((elm: Partial<{instance_type: string}>) => ({ value: elm.instance_type, label: elm.instance_type }))
    return {
      options: machineTypes,
      hasMore: additional && totalPages > additional.page,
      additional: { page: additional ? additional.page + 1 : 1 }
    }
  }

  const getRegion = async () => {
    const res = await Api.Provider.listMetadata('aws', 'regions', '24', 1, region)
    
    setRegionOptions(res.map((elm: any) => ({ value: elm, label: elm })))
  }

  const getFlavors = async () => {
    const res = await Api.Provider.listMetadata('aws', 'supported_flavors', '1', 1, region)
    type apiOption = {
      name: string;
      kubernetes_version: string[];
    };
    
    type apiResponse = apiOption[];
    
    const parse = (data: apiResponse): newStruct => {
      return data.reduce<newStruct>((acc, curr) => {
        acc[curr.name] = {
          selectOptions: curr.kubernetes_version.map((ver) => ({
            label: ver,
            value: ver,
          })),
        };
    
        return acc;
      }, {});
    };

    const parseData = parse(res)
    setFlavorOptions(Object.keys(parseData).map(elm => ({ value: elm, label: elm })))
    setK8sOptions(parseData)
  }

  const getKeys = async () => {
    const res = await Api.Provider.listMetadata('aws', 'ssh_keys', '1', 1, region)
    setSshKeyOptions(res.map((elm: string) => ({ value: elm, label: elm })))
  }

  const getCpu: LoadOptions<OptionType, { page: number }> = async (value, loadedOptions, additional: any) => {
    const res = await Api.Provider.listMetadata('aws', 'machine_types', '15', additional ? additional.page : 1, region)
    const totalPages = res.TotalPages
    const cpu = res.MachineTypes.map((elm: Partial<{vcpus: string}>) => ({ value: elm.vcpus, label: elm.vcpus }))
    return {
      options: cpu,
      hasMore: additional && totalPages > additional.page,
      additional: { page: additional ? additional.page + 1 : 1 }
    }
  }

  const getMem: LoadOptions<OptionType, { page: number }> = async (value, loadedOptions, additional: any) => {
    const res = await Api.Provider.listMetadata('aws', 'machine_types', '15', additional ? additional.page : 1, region)
    const totalPages = res.TotalPages
    const cpu = res.MachineTypes.map((elm: Partial<{memory: string}>) => ({ value: elm.memory, label: elm.memory }))
    return {
      options: cpu,
      hasMore: additional && totalPages > additional.page,
      additional: { page: additional ? additional.page + 1 : 1 }
    }
  }

  useEffect(() => {
    getSecrets()
    getProviders()
    getRegion()
    getFlavors()
    getKeys()
  }, [])

  //inputs
  const formCluster = (e: React.FormEvent<HTMLInputElement>) => {
    setClusterName(e.currentTarget.value)
  }

  const formNamespace = (e: React.FormEvent<HTMLInputElement>) => {
    setNamespace(e.currentTarget.value)
  }

  const formAccessKey = (e: React.FormEvent<HTMLInputElement>) => {
    setAccesskey(e.currentTarget.value)
  }

  const formSecret = (e: React.FormEvent<HTMLInputElement>) => {
    setSecret(e.currentTarget.value)
  }

  const formReplica = (e: React.FormEvent<HTMLInputElement>) => {
    setReplicas(parseInt(e.currentTarget.value) || 0)
  }

  const formReplicaWorkers = (e: React.FormEvent<HTMLInputElement>) => {
    setReplicasWorkers(parseInt(e.currentTarget.value) || 0)
  }

  //selects
  const formProvider = (value: string) => {
    setProvider(value)
  }

  const formSshKey = (value: string) => {
    setSshKey(value)
  }

  const formRegion = (value: string) => {
    setRegion(value)
  }

  const formFlavor = (value: string) => {
    setFlavor(value)
  }

  const formCpu = (option: OptionType | null) => {
    setCpu(option)
  }

  const formMem = (option: OptionType | null) => {
    setMemory(option)
  }

  const formK8s = (value: string) => {
    setK8sVersion(value)
  }

  const formMachineTypes = (option: OptionType | null) => {
    setMachineTypes(option)
  }

  const formCpuWorkers = (option: OptionType | null) => {
    setCpuWorkers(option)
  }

  const formMemWorkers = (option: OptionType | null) => {
    setMemoryWorkers(option)
  }

  const formMachineTypesWorkers = (option: OptionType | null) => {
    setMachineTypesWorkers(option)
  }

  return (
    <>
    <header>
      <h3 className="title"><i className='wizard-text'>Wizard</i> <span>{body.title}</span> {body.ndTitle}</h3>
      <i onClick={handleClose} className="icon-close" />
    </header>
      <div className='box'>
        <Steps handleAction={() => handleAction()}>
          <>
            <h3 className="title-box">Cluster</h3>
            <form className='create-cluster'>
              <Input value={clusterName} onChange={formCluster} type='text' label='cluster name' />
              <Input value={namespace} onChange={formNamespace} type='text' label='namespace' />
              <div className='select-flex'>
                <Select value={provider} onChange={formProvider} options={providerOptions} label='select provider' />
                <Select options={regionOptions} value={region} onChange={formRegion} label='default region' />
              </div>
              <Input disabled type='text' label='secret access ID' value={accessKey} onChange={formAccessKey} />
              <Input disabled type='text' label='secret access key' value={secret} onChange={formSecret} />
              <Input disabled type='text' label='session token' value='' onChange={() => console.log('aa')} />
            </form>
          </>

          <>
            <h3 className="title-box">Infrastructure provider</h3>
            <form className='infra-provider'>
                <Select value={provider} onChange={formProvider} options={providerOptions} label='provider' />
                <Select value={flavor} onChange={formFlavor} options={flavorOptions} label='flavor' />
                <Select options={regionOptions} value={region} onChange={formRegion} label='region' />
                <Select value={k8sVersion} onChange={formK8s} options={k8sOptions?.[flavor]?.selectOptions ?? []} label='kubernetes version' />
                <Select value={sshKey} onChange={formSshKey} options={sshKeyOptions} label='sshKey' />
            </form>
          </>

          <>
            <h3 className="title-box">Control plane</h3>
              <div className='control-plane'>
                  <div className='input-container'>
                    <Input value={replicas} onChange={formReplica} type='text' label='replicas' />
                    <AsyncSelect value={cpu} onChange={formCpu} loadOptions={getCpu} label='CPU' />
                    <AsyncSelect value={memory} onChange={formMem} loadOptions={getMem} label='mem' />
                    <AsyncSelect value={machineTypes} onChange={formMachineTypes} loadOptions={getMachineTypes} label='machineType' />
                  </div>

                  <div className='workers'>
                    <h3 className="title-box">Workers</h3>
                    <Toggle label='InfraNode' value={infraNode} onChange={() => setInfraNode(!infraNode)} />
                    <div className='input-container'> 
                      <Input type='text' label='replicas' value={replicasWorkers} onChange={formReplicaWorkers} />
                      <AsyncSelect value={cpuWorkers} onChange={formCpuWorkers} loadOptions={getCpu} label='CPU' />
                      <AsyncSelect value={memoryWorkers} onChange={formMemWorkers} loadOptions={getMem} label='mem' />
                      <AsyncSelect value={machineTypesWorkers} onChange={formMachineTypesWorkers} loadOptions={getMachineTypes} label='machineType' />
                      <div className='button-container'>
                        <Button onClick={() => createWorkers()} type='gray' size='small' children='Add' />
                      </div>
                    </div>

                    <ul>
                      {(workers || []).map((elm, i = 0) => {
                        return (
                          <li key={elm.id}>
                            <p>{clusterName}-mp-{i}</p>
                            <i onClick={() => deleteWorkers(elm.id)} className='icon-close' />
                          </li>
                        )
                      })}
                    </ul>
                  </div>
              </div>
          </>
        </Steps>
      </div>
  </>
  )
}

export default ClusterWizard