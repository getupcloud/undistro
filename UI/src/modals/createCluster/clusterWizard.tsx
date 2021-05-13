/* eslint-disable react-hooks/exhaustive-deps */
import React, { FC, useEffect, useState } from 'react'
import store from '../store'
import Input from '@components/input'
import Select from '@components/select'
import Modals from 'util/modals'
import Steps from './steps'
import Button from '@components/button'
import Api from 'util/api'

type Props = {
  handleClose: () => void
}


const ClusterWizard: FC<Props> = ({ handleClose }) => {
  const body = store.useState((s: any) => s.body)
  const [accessKey, setAccesskey] = useState<any>('')
  const [secret, setSecret] = useState<any>('')
  const [region, setRegion] = useState<any>('')
  const [clusterName, setClusterName] = useState<any>()
  const [namespace, setNamespace] = useState<any>('')
  const [provider, setProvider] = useState<any>('')
  const [flavor, setFlavor] = useState<string>('')
  const [k8sVersion, setK8sVersion] = useState<string>('')
  const [replicas, setReplicas] = useState<any>('')
  const [machineTypes, setMachineTypes] = useState<any>('')
  const machineOptions = [{ value: 't3.medium', label: 't3.medium'}]
  const flavorOptions = [{ value: 'eks', label: 'EKS'}, { value: 'ec2', label: 'EC2'}]
  const providerOptions = [{ value: provider, label: 'aws' }]
  const regionOptions = [{ value: region, label: 'us-east-1'}]
  const k8sOptions = [{ value: 'v1.18.9', label: 'v1.18.9'}]
  // const handleAction = () => {
  //   handleClose()
  //   if (body.handleAction) body.handleAction()
  // }

  const showModal = () => {
    handleClose()
    Modals.show('create-cluster', {
      title: 'Create',
			ndTitle: 'Cluster'
    })
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

  return (
    <>
    <header>
      <h3 className="title"><span>{body.title}</span> {body.ndTitle}</h3>
      <i onClick={handleClose} className="icon-close" />
    </header>
      <div className='box'>
        <Steps handleAction={() => console.log('children')}>
          <>
            <h3 className="title-box">Cluster</h3>
            <form className='create-cluster'>
              <Input value={clusterName} onChange={setClusterName} type='text' label='cluster name' />
              <Input value={namespace} onChange={setNamespace} type='text' label='namespace' />
              <div className='select-flex'>
                <Select value={provider} onChange={setProvider} options={providerOptions} label='select provider' />
                <Select options={regionOptions} value={region} onChange={setRegion} label='default region' />
              </div>
              <Input disabled type='text' label='secret access ID' value={accessKey} onChange={setAccesskey} />
              <Input disabled type='text' label='secret access key' value={secret} onChange={setSecret} />
              <Input disabled type='text' label='session token' value='' onChange={() => console.log('aa')} />
            </form>
          </>

          <>
            <h3 className="title-box">Infrastructure provider</h3>
            <form className='infra-provider'>
                <Select value={provider} onChange={setProvider} options={providerOptions} label='provider' />
                <Select value={flavor} onChange={setFlavor} options={flavorOptions} label='flavor' />
                <Select options={regionOptions} value={region} onChange={setRegion} label='region' />
                <Select value={k8sVersion} onChange={setK8sVersion} options={k8sOptions} label='kubernetes version' />
                <Select label='sshKey' />
            </form>
          </>

          <>
            <h3 className="title-box">Control plane</h3>
              <form className='control-plane'>
                  <div className='input-container'>
                    <Input value={replicas} onChange={setReplicas} type='text' label='replicas' />
                    <Select label='CPU' />
                    <Select label='mem' />
                    <Select value={machineTypes} onChange={setMachineTypes} options={machineOptions} label='machineType' />
                  </div>

                  <div className='workers'>
                    <h3 className="title-box">Workers</h3>
                    <div className='input-container'>
                      <Input type='text' label='replicas' value='' onChange={() => console.log('aa')} />
                      <Select label='CPU' />
                      <Select label='mem' />
                      <Select value={machineTypes} onChange={setMachineTypes} options={machineOptions} label='machineType' />
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