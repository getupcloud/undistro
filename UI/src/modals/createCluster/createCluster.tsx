import React, { FC } from 'react'
import store from '../store'
import Wizard from '@assets/images/wizard.svg'
import Advanced from '@assets/images/advanced.svg'
import Modals from 'util/modals'

type Props = {
  handleClose: () => void
}

const CreateCluster: FC<Props> = ({ handleClose }) => {
  const body = store.useState((s: any) => s.body)
  // const handleAction = () => {
  //   handleClose()
  //   if (body.handleAction) body.handleAction()
  // }

  const wizardModal = () => {
    handleClose()
    Modals.show('wizard-cluster', {
      title: 'Create',
			ndTitle: 'Cluster',
      width: '720',
      height: '600'
    })
  }

  const advancedModal = () => {
    handleClose()
    Modals.show('advanced-cluster', {
      title: 'Create',
			ndTitle: 'Cluster',
      width: '720',
      height: '600'
    })
  }


  return (
    <>
    <header>
      <h3 className="title"><span>{body.title}</span> {body.ndTitle}</h3>
      <i onClick={handleClose} className="icon-close" />
    </header>
    <div className='box-create'>
      <section>
        <div onClick={() => wizardModal()} className='option'>
          <div className='square'>
            <img src={Wizard} alt="wizard" />
          </div>
          <div className='text'>
            <h1>Wizard</h1>
            <p>The fastest way to start.</p>
            <p>Create a cluster in just a few steps.</p>
          </div>
        </div>

        <div onClick={() => advancedModal()} className='option'>
          <div className='square'>
            <img src={Advanced} alt="wizard" />
          </div>
          <div className='text'>
            <h1>Advanced</h1>
            <p>No shortcuts.</p> 
            <p>Control every aspect of the cluster.</p>
          </div>
        </div>
      </section>
    </div>
  </>
  )
}

export default CreateCluster