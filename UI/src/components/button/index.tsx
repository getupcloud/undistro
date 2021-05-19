import React, { FC } from 'react'
import Classnames from 'classnames'
import './index.scss'

type Props = {
  children?: string,
  type: string,
  size: string,
  disabled?: boolean,
  onClick?: () => void
}

const Button: FC<Props> = ({ 
  children, 
  type, 
  size, 
  disabled,
  onClick 
}) => {
  const style = Classnames('button',
    `button--${type}`,
    `button--${size}`,
  )

  return (
    <button 
      className={style} 
      disabled={disabled} 
      onClick={onClick}
    >
      {children}
    </button>
  )
}

Button.defaultProps = {
  type: 'primary',
}

export default Button