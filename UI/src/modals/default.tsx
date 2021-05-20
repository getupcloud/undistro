import React from 'react'
import Button from '@components/button'
import store from './store'

type Props = {
  handleClose: () => void
}

function DefaultModal({ handleClose }: Props) {
  const body = store.useState((s: any) => s.body)
  const handleAction = () => {
    handleClose()
    if (body.handleAction) body.handleAction()
  }

  return (
    <>
      <header>
        <h3 className="title"><span>{body.title}</span> {body.ndTitle}</h3>
        <i onClick={handleClose} className="icon-close" />
      </header>
      <section>
        <p>testando</p>
      </section>
      <footer>
        <Button type='primary' size='medium' onClick={handleAction} children='next' />
      </footer>
    </>
  )
}

export default DefaultModal
