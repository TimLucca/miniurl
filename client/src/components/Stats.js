import React, {Component} from 'react'
import axios from "axios"
import { Header, Form, Input } from 'semantic-ui-react'

class Stats extends Component {
  constructor(props) {
    super(props)

    this.state = {
      miniurl: '',
      res: '',
      hits: '',
      submitted: false,
    }
  }

  componentDidMount() {
    
  }

  onChange = event => {
    this.setState({
      [event.target.name]: event.target.value
    })
  }

  onSubmit = () => {
    let {miniurl} = this.state 
    miniurl = miniurl.substring(miniurl.length-6, miniurl.length)
    
    if (miniurl) {
      axios.post(
        '/api/current',
        {miniurl},
        {
          headers: {
            'Content-Type': 'application/json'
          }
        }
      ).then(res => {
        this.setState({res: res, submitted: true})
      }).catch(e => {
        if(e.response) {
          this.setState({res: e.response, submitted: true})
        }
        else {
          this.setState({submitted: true})
        }
      })
    }
  }

  isMini = () => {
    let res = this.state.res
    if(!this.state.submitted){
      return
    }
    if (res.status === 200) {
      console.log(res.data)
      return (
        <div>
          <p><strong>MiniURL: </strong><a href={res.data.miniurl}>{res.data.miniurl}</a></p>
          <p><strong>Original URL: </strong><a href={res.data.long}>{res.data.long}</a></p>
          <p><strong>Hits: </strong>{res.data.hits}</p>
        </div>
        
      )
    }else if (res.status === 404) {
      return (
        <div className='ui negative message'>
          <i className='close icon'></i>
          <div className='header'>
            Could not find that MiniURL
          </div>
        </div>
      )
    } else if(!this.state.submitted){
      return
    } else {
      return (
        <div className='ui negative message'>
          <i className='close icon'></i>
          <div className='header'>
            An error occured
          </div>
        </div>
      )
    }
  }

  render() {
    const styles = {
      paddingBottom: '2em',
    }

    return (
      <div>
        <div className='row' style={styles}>
          <Header className='header' as='h1'>
            MiniURL Statistics
          </Header>
        </div>
        <div className='row' style={styles}>
          <Form onSubmit={this.onSubmit}>
            <Input type='text' name='miniurl' value={this.state.miniurl} onChange={this.onChange} fluid placeholder='Enter miniURL' />
          </Form>
        </div>
        <button onClick={this.onSubmit}>Get Stats</button>
        {this.isMini()}
      </div>
    )
  }
}

export default Stats;