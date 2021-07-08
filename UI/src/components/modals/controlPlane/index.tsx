import React, { FC } from 'react'
import AsyncSelect from '@components/asyncSelect'
import Input from '@components/input'
import Button from '@components/button'
import Toggle from '@components/toggle'
import { TypeOption, TypeControlPlane } from '../../../types/cluster'

const ControlPlane: FC<TypeControlPlane> = ({
  replicas,
  setReplicas,
  cpu,
  setCpu,
  getCpu,
  getMem,
  memory,
  setMemory,
  machineTypes,
  setMachineTypes,
  getMachineTypes,
  infraNode,
  setInfraNode,
  createWorkers,
  deleteWorkers,
  workers,
  replicasWorkers,
  setReplicasWorkers,
  cpuWorkers,
  setCpuWorkers,
  memoryWorkers,
  setMemoryWorkers,
  machineTypesWorkers,
  setMachineTypesWorkers,
  clusterName
}) => {

  const formReplica = (e: React.FormEvent<HTMLInputElement>) => {
    setReplicas(parseInt(e.currentTarget.value) || 0)
  }

  const formReplicaWorkers = (e: React.FormEvent<HTMLInputElement>) => {
    setReplicasWorkers(parseInt(e.currentTarget.value) || 0)
  }

  const formCpu = (option: TypeOption | null) => {
    setCpu(option)
  }

  const formMem = (option: TypeOption | null) => {
    setMemory(option)
  }

  const formMachineTypes = (option: TypeOption | null) => {
    setMachineTypes(option)
  }

  const formCpuWorkers = (option: TypeOption | null) => {
    setCpuWorkers(option)
  }

  const formMemWorkers = (option: TypeOption | null) => {
    setMemoryWorkers(option)
  }

  const formMachineTypesWorkers = (option: TypeOption | null) => {
    setMachineTypesWorkers(option)
  }


  return (
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
  )
}

export default ControlPlane