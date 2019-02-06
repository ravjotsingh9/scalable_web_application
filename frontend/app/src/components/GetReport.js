import React, { Component } from "react";
import { Label, Table, Pagination, Button } from "react-bootstrap";
import axios from "axios";
import { JsonToTable } from "react-json-to-table";

class GetReport extends Component {
  constructor(props) {
    super(props);
    this.state = {
      contents: [],
      api_getReport_address: ""
    };
  }
  componentDidMount() {
    this.props.getHandler();
  }

  render() {
    let _this = this;
    return (
      <div>
        {/* <Button onClick={() => this.getContent()}>Get</Button> */}

        <h3>Report Summary</h3>

        {/* {this.state.contents.length} */}
        <Table striped bordered condensed hover>
          <thead>
            <tr>
              <th>Employee ID</th>
              <th>Pay Period</th>
              <th>Amount Paid</th>
            </tr>
          </thead>

          {/* <JsonToTable json={rows} /> */}
          <tbody>
            {_this.props.contents.map(function(item, key) {
              return (
                <tr key={key}>
                  <td>{item.employeeID}</td>
                  <td>{item.payDate}</td>
                  <td>{item.salary}</td>
                </tr>
              );
            })}
          </tbody>
        </Table>
      </div>
    );
  }
}

export default GetReport;
