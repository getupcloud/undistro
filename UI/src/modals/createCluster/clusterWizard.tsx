/* eslint-disable react-hooks/exhaustive-deps */
import React, { FC, useEffect, useState } from 'react'
import store from '../store'
import Input from '@components/input'
import Select from '@components/select'
// import Modals from 'util/modals'
import Steps from './steps'
import Button from '@components/button'
import Api from 'util/api'

type Props = {
  handleClose: () => void
}


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
  const [machineTypes, setMachineTypes] = useState<string>('')
  const [replicasWorkers, setReplicasWorkers] = useState<number>(0)
  const [machineTypesWorkers, setMachineTypesWorkers] = useState<string>('')
  const machineOptions = [{ value: 't3.medium', label: 't3.medium'}]
  const flavorOptions = [{ value: 'eks', label: 'EKS'}, { value: 'ec2', label: 'EC2'}]
  const providerOptions = [{ value: provider, label: 'aws' }]
  const regionOptions = [{ value: region, label: 'us-east-1'}]
  const k8sOptions = [{ value: 'v1.18.9', label: 'v1.18.9'}]
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

  const handleAction = () => {
    const cluster = {
      "name": clusterName,
      "namespace": namespace
    }

    const spec = {
      "kubernetesVersion": k8sVersion,
      "controlPlane": {
        "machineTypes": machineTypes,
        "replicas": replicas
      },

      "infrastructureProvider": {
        "flavor": flavor,
        "name": provider,
        "region": region
      },

      "workers": [{
        "machineTypes": machineTypes,
        "replicas": replicas
      }]
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

    Api.Cluster.post(data)
      .then(res => console.log(res, 'success'))
    
    Api.Cluster.postPolicies(dataPolicies)
      .then(res => console.log(res, 'success'))
  }

  const getSecrets = () => {
    Api.Secret.list()
      .then(res => {
        setAccesskey(atob(res.data.accessKeyID))
        setSecret(atob(res.data.secretAccessKey))
        setRegion(atob(res.data.region))
      })
  }

  const getProviders = () => {
    Api.Provider.list()
      .then(res => {
        setProvider(res.items[0].metadata.name)
      })
  }

  useEffect(() => {
    getSecrets()
    getProviders()
    Api.Cluster.list()
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

  const formRegion = (value: string) => {
    setRegion(value)
  }

  const formFlavor = (value: string) => {
    setFlavor(value)
  }

  const formK8s = (value: string) => {
    setK8sVersion(value)
  }

  const formMachineTypes = (value: string) => {
    setMachineTypes(value)
  }

  const formMachineTypesWorkers = (value: string) => {
    setMachineTypesWorkers(value)
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
                <Select value={k8sVersion} onChange={formK8s} options={k8sOptions} label='kubernetes version' />
                <Input type='text' value='' onChange={() => console.log('aa')} label='sshKey' />
            </form>
          </>

          <>
            <h3 className="title-box">Control plane</h3>
              <form className='control-plane'>
                  <div className='input-container'>
                    <Input value={replicas} onChange={formReplica} type='text' label='replicas' />
                    {/* <Select label='CPU' /> */}
                    {/* <Select label='mem' /> */}
                    <Select value={machineTypes} onChange={formMachineTypes} options={machineOptions} label='machineType' />
                  </div>

                  <div className='workers'>
                    <h3 className="title-box">Workers</h3>
                    <div className='input-container'>
                      <Input type='text' label='replicas' value={replicasWorkers} onChange={formReplicaWorkers} />
                      {/* <Select label='CPU' /> */}
                      {/* <Select label='mem' /> */}
                      <Select value={machineTypesWorkers} onChange={formMachineTypesWorkers} options={machineOptions} label='machineType' />
                      <div className='button-container'>
                        <Button type='gray' size='small' children='Add' />
                      </div>
                    </div>

                    <ul>
                      <li>
                        <p>clusterName-mp-0</p>
                        <i className='icon-close' />
                      </li>
                      <li>
                        <p>clusterName-mp-1</p>
                        <i className='icon-close' />
                      </li>
                      <li>
                        <p>clusterName-mp-2</p>
                        <i className='icon-close' />
                      </li>
                      <li>
                        <p>clusterName-mp-3</p>
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

export default ClusterWizard