import React, { FC, ReactChild, useState } from 'react'
import Button from '@components/button'

type Props = {
  children: ReactChild[],
  handleAction: () => void
}

const Steps: FC<Props> = ({ children, handleAction }) => {
  const [index, setIndex] = useState<number>(0)

  if(index === 3) {
    handleAction()
  }


  return (
    <>
      {children[index]}

      <footer>
        <Button disabled={index === 0} onClick={() => setIndex(index - 1)} type='black' size='large' children='back' />
        <Button onClick={() => setIndex(index + 1)} type='primary' size='large' children='next'/>
      </footer>
    </>
  )
}

export default Steps