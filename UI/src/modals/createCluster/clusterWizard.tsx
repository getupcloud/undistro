/* eslint-disable jsx-a11y/no-access-key */
/* eslint-disable react-hooks/exhaustive-deps */
import React, { FC, useEffect, useState } from 'react'
import store from '../store'
import { LoadOptions } from "react-select-async-paginate";
import CreateCluster from '@components/modals/cluster'
import Infra from '@components/modals/infrastructureProvider'
import ControlPlane from '@components/modals/controlPlane'
import { generateId } from 'util/helpers'
import Steps from './steps'
import Api from 'util/api'
import { TypeOption, TypeWorker, TypeSelectOptions } from '../../types/cluster'
import { TypeModal } from '../../types/generic'

const ClusterWizard: FC<TypeModal> = ({ handleClose }) => {
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
  const [machineTypes, setMachineTypes] = useState<TypeOption | null>(null)
  const [memory, setMemory] = useState<TypeOption | null>(null)
  const [cpu, setCpu] = useState<TypeOption | null>(null)
  const [replicasWorkers, setReplicasWorkers] = useState<number>(0)
  const [memoryWorkers, setMemoryWorkers] = useState<TypeOption | null>(null)
  const [cpuWorkers, setCpuWorkers] = useState<TypeOption | null>(null)
  const [machineTypesWorkers, setMachineTypesWorkers] = useState<TypeOption | null>(null)
  const [regionOptions, setRegionOptions] = useState<[]>([])
  const [flavorOptions, setFlavorOptions] = useState<TypeOption[]>([])
  const [k8sOptions, setK8sOptions] = useState<TypeSelectOptions>()
  const [sshKey, setSshKey] = useState<string>('')
  const [sshKeyOptions, setSshKeyOptions] = useState<string[]>([])
  const providerOptions = [{ value: provider, label: 'aws' }]

  const handleAction = () => {
    const getWorkers = workers.map(elm => ({ machineType: elm.machineType, replicas: elm.replicas, infraNode: elm.infraNode }))

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

  const getMachineTypes: LoadOptions<TypeOption, { page: number }> = async (value, loadedOptions, additional: any) => {
    const res = await Api.Provider.listMetadata('aws', 'machineTypes', '15', additional ? additional.page : 1, region)
    const totalPages = res.TotalPages
    const machineTypes = res.MachineTypes.map((elm: Partial<{ instanceType: string }>) => ({ value: elm.instanceType, label: elm.instanceType }))
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
    const res = await Api.Provider.listMetadata('aws', 'supportedFlavors', '1', 1, region)
    type apiOption = {
      name: string;
      kubernetesVersion: string[];
    };

    type apiResponse = apiOption[];

    const parse = (data: apiResponse): TypeSelectOptions => {
      return data.reduce<TypeSelectOptions>((acc, curr) => {
        acc[curr.name] = {
          selectOptions: curr.kubernetesVersion.map((ver) => ({
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
    const res = await Api.Provider.listMetadata('aws', 'sshKeys', '1', 1, region)
    setSshKeyOptions(res.map((elm: string) => ({ value: elm, label: elm })))
  }

  const getCpu: LoadOptions<TypeOption, { page: number }> = async (value, loadedOptions, additional: any) => {
    const res = await Api.Provider.listMetadata('aws', 'machineTypes', '15', additional ? additional.page : 1, region)
    const totalPages = res.TotalPages
    const cpu = res.MachineTypes.map((elm: Partial<{ vcpus: string }>) => ({ value: elm.vcpus, label: elm.vcpus }))
    return {
      options: cpu,
      hasMore: additional && totalPages > additional.page,
      additional: { page: additional ? additional.page + 1 : 1 }
    }
  }

  const getMem: LoadOptions<TypeOption, { page: number }> = async (value, loadedOptions, additional: any) => {
    const res = await Api.Provider.listMetadata('aws', 'machineTypes', '15', additional ? additional.page : 1, region)
    const totalPages = res.TotalPages
    const cpu = res.MachineTypes.map((elm: Partial<{ memory: string }>) => ({ value: elm.memory, label: elm.memory }))
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

  return (
    <>
      <header>
        <h3 className="title"><i className='wizard-text'>Wizard</i> <span>{body.title}</span> {body.ndTitle}</h3>
        <i onClick={handleClose} className="icon-close" />
      </header>
      <div className='box'>
        <Steps handleAction={() => handleAction()}>
          <CreateCluster 
            clusterName={clusterName}
            setClusterName={setClusterName}
            namespace={namespace}
            setNamespace={setNamespace}
            provider={provider}
            setProvider={setProvider}
            providerOptions={providerOptions}
            region={region}
            setRegion={setRegion}
            regionOptions={regionOptions}
            accessKey={accessKey}
            setAccesskey={setAccesskey}
            secret={secret}
            setSecret={setSecret}
          />

          <Infra 
            provider={provider}
            setProvider={setProvider}
            providerOptions={providerOptions}
            flavor={flavor}
            setFlavor={setFlavor}
            flavorOptions={flavorOptions}
            region={region}
            setRegion={setRegion}
            regionOptions={regionOptions}
            k8sVersion={k8sVersion}
            setK8sVersion={setK8sVersion}
            k8sOptions={k8sOptions}
            sshKey={sshKey}
            setSshKey={setSshKey}
            sshKeyOptions={sshKeyOptions}
          />

          <ControlPlane 
            replicas={replicas}
            setReplicas={setReplicas}
            cpu={cpu}
            setCpu={setCpu}
            getCpu={getCpu}
            memory={memory}
            setMemory={setMemory}
            getMem={getMem}
            machineTypes={machineTypes}
            setMachineTypes={setMachineTypes}
            getMachineTypes={getMachineTypes}
            infraNode={infraNode}
            setInfraNode={setInfraNode}
            workers={workers}
            createWorkers={createWorkers}
            deleteWorkers={deleteWorkers}
            replicasWorkers={replicasWorkers}
            setReplicasWorkers={setReplicasWorkers}
            cpuWorkers={cpuWorkers}
            setCpuWorkers={setCpuWorkers}
            memoryWorkers={memoryWorkers}
            setMemoryWorkers={setMemoryWorkers}
            machineTypesWorkers={machineTypesWorkers}
            setMachineTypesWorkers={setMachineTypesWorkers}
            clusterName={clusterName}
          />

        </Steps>
      </div>
    </>
  )
}

export default ClusterWizard