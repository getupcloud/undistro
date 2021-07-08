/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable jsx-a11y/no-access-key */
import React, { FC, useState, useEffect } from 'react'
import store from '../store'
import Input from '@components/input'
import CreateCluster from '@components/modals/cluster'
import Infra from '@components/modals/infrastructureProvider'
import Steps from './steps'
import Button from '@components/button'
import Api from 'util/api'
import Toggle from '@components/toggle'
import { TypeOption, TypeSelectOptions } from '../../types/cluster'
import { TypeModal } from '../../types/generic'

const ClusterAdvanced: FC<TypeModal> = ({ handleClose }) => {
  const body = store.useState((s: any) => s.body)
  const [accessKey, setAccesskey] = useState<string>('')
  const [secret, setSecret] = useState<string>('')
  const [region, setRegion] = useState<string>('')
  const [clusterName, setClusterName] = useState<string>('')
  const [namespace, setNamespace] = useState<string>('')
  const [provider, setProvider] = useState<string>('')
  const [flavor, setFlavor] = useState<string>('')
  const [k8sVersion, setK8sVersion] = useState<string>('')
  // const [replicas, setReplicas] = useState<number>(0)
  // const [infraNode, setInfraNode] = useState<boolean>(false)
  // const [workers, setWorkers] = useState<TypeWorker[]>([])
  // const [machineTypes, setMachineTypes] = useState<TypeOption | null>(null)
  // const [memory, setMemory] = useState<TypeOption | null>(null)
  // const [cpu, setCpu] = useState<TypeOption | null>(null)
  // const [replicasWorkers, setReplicasWorkers] = useState<number>(0)
  // const [memoryWorkers, setMemoryWorkers] = useState<TypeOption | null>(null)
  // const [cpuWorkers, setCpuWorkers] = useState<TypeOption | null>(null)
  // const [machineTypesWorkers, setMachineTypesWorkers] = useState<TypeOption | null>(null)
  const [regionOptions, setRegionOptions] = useState<[]>([])
  const [flavorOptions, setFlavorOptions] = useState<TypeOption[]>([])
  const [k8sOptions, setK8sOptions] = useState<TypeSelectOptions>()
  const [sshKey, setSshKey] = useState<string>('')
  const [sshKeyOptions, setSshKeyOptions] = useState<string[]>([])
  const providerOptions = [{ value: provider, label: 'aws' }]
	const [test, setTest] = useState(false)
  
  // const handleAction = () => {
  //   handleClose()
  //   if (body.handleAction) body.handleAction()
  // }

  // const showModal = () => {
  //   handleClose()
  //   Modals.show('create-cluster', {
  //     title: 'Create',
	// 		ndTitle: 'Cluster'
  //   })
  // }


  const getSecrets = () => {
    Api.Secret.list()
      .then(res => {
        setAccesskey(atob(res.data.accessKeyID))
        setSecret(atob(res.data.secretAccessKey))
        setRegion(atob(res.data.region))
      })
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

  useEffect(() => {
    getSecrets()
    getRegion()
    getFlavors()
    getKeys()
  }, [])



  return (
    <>
    <header>
      <h3 className="title"><span>{body.title}</span> {body.ndTitle}</h3>
      <i onClick={handleClose} className="icon-close" />
    </header>
      <div className='box'>
        <Steps handleAction={() => console.log('test')}>
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
          <>
            <h3 className="title-box">infra network - VPC</h3>
            <form className='infra-network'>
              <div className='input-container'>
                {/* <Select options={regionOptions} value={region} onChange={formRegion} label='ID' /> */}
                <Input type='text' label='CIDR block' value='' onChange={() => console.log('aa')} />
              </div>

              <div className='subnet'>
                <h3 className="title-box">subnet</h3>
                
                <Toggle label='Is public' value={test} onChange={() => setTest(!test)} />
                <div className='subnet-inputs'>
                  {/* <Select options={regionOptions} value={region} onChange={formRegion} label='ID' /> */}
                  <Input type='text' label='zone' value='' onChange={() => console.log('aa')} />
                  <Input type='text' label='CIDR block' value='' onChange={() => console.log('aa')} />
                  <div className='button-container'>
                    <Button onClick={() => console.log('test')} type='gray' size='small' children='Add' />
                  </div>
                </div>

                <ul>
                  <li>
                    <p>allowedBlock-0</p>
                    <i className='icon-close' />
                  </li>
                  <li>
                    <p>allowedBlock-1</p>
                    <i className='icon-close' />
                  </li>
                  <li>
                    <p>allowedBlock-2</p>
                    <i className='icon-close' />
                  </li>
                  <li>
                    <p>allowedBlock-3</p>
                    <i className='icon-close' />
                  </li>
                </ul>
              </div>
            </form>
          </>
        </Steps>
      </div>
  </>
  )
}

export default ClusterAdvanced

/* <>
<h3 className="title-box">infra network - VPC</h3>
<form className='infra-network'>
  <div className='input-container'>
    <Select label='ID' />
    <Input type='text' label='CIDR block' value='' onChange={() => console.log('aa')} />
  </div>

  <div className='subnet'>
    <h3 className="title-box">subnet</h3>
    
    <Toggle label='Is public' value={test} onChange={() => setTest(!test)} />
    <div className='subnet-inputs'>
      <Select label='ID' />
      <Input type='text' label='zone' value='' onChange={() => console.log('aa')} />
      <Input type='text' label='CIDR block' value='' onChange={() => console.log('aa')} />
      <div className='button-container'>
        <Button type='gray' size='small' children='Add' />
      </div>
    </div>

    <ul>
      <li>
        <p>allowedBlock-0</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>allowedBlock-1</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>allowedBlock-2</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>allowedBlock-3</p>
        <i className='icon-close' />
      </li>
    </ul>
  </div>
</form>
</>

<>
<form>
  <Input type='text' label='API server port' value='' onChange={() => console.log('aa')} />
  <Input type='text' label='serice domain' value='' onChange={() => console.log('aa')} />
  <div className='input-flex'>
    <Input type='text' label='pods ranges' value='' onChange={() => console.log('aa')} />
    <Input type='text' label='service ranges' value='' onChange={() => console.log('aa')} />
  </div>
  <Select label='CNI plugin' />

  <div className='flags-container'>
    <Input type='text' label='flags' value='' onChange={() => console.log('aa')} />

    <ul>
      <li>
        <p>flag-0</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>flag-1</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>flag-2</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>flag-3</p>
        <i className='icon-close' />
      </li>
    </ul>
  </div>
</form>
</>

<>
<form>
  <Toggle label='enabled' value={test} onChange={() => setTest(!test)} />
  <Toggle label='disable ingress rules' value={test} onChange={() => setTest(!test)} />
  <div className='flex-text'>
    <p>user default blocks CIDR</p>
    <span>198.51.100.2</span>
  </div>

  <div className='input-container'>
    <Input type='text' label='replicas' value='' onChange={() => console.log('aa')} />
    <Select label='CPU' />
    <Select label='mem' />
    <Select label='machineType' />
  </div>

  <div className='flags-container'>
    <Input type='text' label='allowed blocks CIDR' value='' onChange={() => console.log('aa')} />

    <ul>
      <li>
        <p>allowedBlock-0</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>allowedBlock-1</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>allowedBlock-2</p>
        <i className='icon-close' />
      </li>
      <li>
        <p>allowedBlock-3</p>
        <i className='icon-close' />
      </li>
    </ul>
  </div>
</form>
</>

<>
<form className='control-plane'>
  <div className='input-container'>
    <Input type='text' label='replicas' value='' onChange={() => console.log('aa')} />
    <Select label='CPU' />
    <Select label='mem' />

</> */