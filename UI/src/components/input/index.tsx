import React, { FC, FormEventHandler } from 'react'
import Classnames from 'classnames'
import './index.scss'


type Props = {
  type: string,
  label?: string,
  placeholder?: string,
  value: string | number,
  disabled?: boolean,
  validator?: {},
  onChange: FormEventHandler<HTMLInputElement>,
}

const Input: FC<Props> = ({
  type,
  label,
  placeholder,
  value,
  disabled,
  validator,
  onChange
}) => {
  const style = Classnames('input', {
    'input--error': validator
  })

  return (
    <div className={style}>
      {label && <label>{label}</label>}
      <input
        type={type}
        value={value}
        placeholder={placeholder}
        disabled={disabled}
        onChange={onChange}
      />
    </div>
  )
}

export default Input