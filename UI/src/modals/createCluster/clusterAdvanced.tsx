import React, { FC, useState } from 'react'
import store from '../store'
import Input from '@components/input'
import Select from '@components/select'
import Modals from 'util/modals'
import Steps from './steps'
import Button from '@components/button'
import Toggle from '@components/toggle'
type Props = {
  handleClose: () => void
}

const ClusterAdvanced: FC<Props> = ({ handleClose }) => {
  const body = store.useState((s: any) => s.body)
	const [test, setTest] = useState(false)
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
              <Input type='text' label='cluster name' value='' onChange={() => console.log('aa')} />
              <Input type='text' label='namespace' value='' onChange={() => console.log('aa')} />
              <div className='select-flex'>
                <Select label='select provider' />
                <Select label='select provider' />
              </div>
              <Input type='text' label='secret access ID' value='' onChange={() => console.log('aa')} />
              <Input type='text' label='secret access key' value='' onChange={() => console.log('aa')} />
              <Input type='text' label='session token' value='' onChange={() => console.log('aa')} />
            </form>
          </>

          <>
            <h3 className="title-box">Infrastructure provider</h3>
            <form className='infra-provider'>
                <Select label='provider' />
                <Select label='flavor' />
                <Select label='region' />
                <Select label='kubernetes version' />
                <Select label='sshKey' />
            </form>
          </>

          <>
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
                <Select label='machineType' />
              </div>
            </form>
          </>
        </Steps>
      </div>
  </>
  )
}

export default ClusterAdvanced