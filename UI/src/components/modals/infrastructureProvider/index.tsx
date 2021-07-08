import React, { FC } from 'react'
import Select from '@components/select'
import { TypeInfra } from '../../../types/cluster'

const InfrastructureProvider: FC<TypeInfra> = ({
  provider,
  setProvider,
  providerOptions,
  flavor,
  setFlavor,
  flavorOptions,
  k8sVersion,
  setK8sVersion,
  k8sOptions,
  regionOptions,
  region,
  setRegion,
  sshKey,
  setSshKey,
  sshKeyOptions  
}) => {

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

  const formK8s = (value: string) => {
    setK8sVersion(value)
  }

  return (
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
  )
}

export default InfrastructureProvider