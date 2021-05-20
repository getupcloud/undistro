import React from 'react'
import store from './store'
import './index.scss'

import Default from './default'
import CreateCluster from './createCluster/createCluster'
import AdvancedCluster from './createCluster/clusterAdvanced'
import WizardCluster from './createCluster/clusterWizard'

function RenderModal() {
  const id = store.useState((s: any) => s.id)

  const handleClose = () => {
    store.update((s: any) => {
      s.show = false
    })
  }

  switch (id) {
    case 'default':
      return <Default handleClose={handleClose} />
    case 'create-cluster':
      return <CreateCluster handleClose={handleClose} />
    case 'wizard-cluster':
      return <WizardCluster handleClose={handleClose} />
    case 'advanced-cluster':
      return <AdvancedCluster handleClose={handleClose} />
    default:
      return null
  }
}

function Modals() {
  const show = store.useState((s:any) => s.show)
  const body = store.useState((s:any) => s.body)
  if (!show) return null

  return (
    <div className="modal-container">
      <div className="modal-content" style={{ width: `${body.width}px`, height: `${body.height}px`}}>
        <RenderModal />
      </div>
    </div>
  )
}

export default Modals