import React, { FC, FormEventHandler } from 'react'
import Classnames from 'classnames'
import './index.scss'

type FallbackType = Element;
type ConstraintType = Element;

type Props<T extends ConstraintType = FallbackType> = {
  type: string,
  label?: string,
  placeholder?: string,
  value: string,
  disabled?: boolean,
  validator?: {},
  onChange: FormEventHandler<T>,
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